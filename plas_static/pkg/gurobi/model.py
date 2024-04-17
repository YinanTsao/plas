import math
from gurobipy import Model, GRB
import os
import sys
import json
from datetime import datetime


snapshot_dir = "../../data/snapshot/"
files = os.listdir(snapshot_dir)

# Filter the list to include only JSON files
json_files = [f for f in files if f.endswith('.json')]

# Sort the list based on the timestamps in the file names
json_files.sort(key=lambda x: datetime.strptime(x, 'fodder_%Y%m%d%H%M%S.json'))

# Load the JSON file with the nearest timestamp
with open(os.path.join(snapshot_dir, json_files[-1]), 'r') as f:
    data = json.load(f)



sites = data["nodes"]
users = data["users"]
capacities = data["capacities"]
latency_node_user = {(i, j): data["latency_node_user"][f"('{i}', '{j}')"] for i in sites for j in users}
service_rate = data["service_rate"] / 1e3 # req/ms
request_rate = data["rps"] / 1e3  # req/ms
slo = data["slo"]  # ms
deployment_name = data["deployment_name"]
opti_pref = data["opti_pref"]


# sites = ['S1', 'S2', 'S3', 'S4']
# users = ['U1', 'U2', 'U3', 'U4', 'U5', 'U6']
# capacities = {'S1': 10, 'S2': 20, 'S3': 17, 'S4': 15}
# latency_node_user = { # Define for all site-user pairs
#     ('S1', 'U1'): 4, ('S1', 'U2'): 5, ('S1', 'U3'): 6, ('S1', 'U4'): 8, ('S1', 'U5'): 7, ('S1', 'U6'): 6,
#     ('S2', 'U1'): 6, ('S2', 'U2'): 4, ('S2', 'U3'): 3, ('S2', 'U4'): 5, ('S2', 'U5'): 6, ('S2', 'U6'): 7,
#     ('S3', 'U1'): 9, ('S3', 'U2'): 7, ('S3', 'U3'): 4, ('S3', 'U4'): 3 ,('S3', 'U5'): 5, ('S3', 'U6'): 6,
#     ('S4', 'U1'): 8, ('S4', 'U2'): 6, ('S4', 'U3'): 5, ('S4', 'U4'): 4, ('S4', 'U5'): 3, ('S4', 'U6'): 2
# } # ms
# service_rate = 10 # req/s
# request_rate = 20 # req/s
# slo = 7 # ms


# Setup the big-M value00
M = slo + max(latency_node_user.values()) + max(capacities.values()) + 1e6 # Upper bound for the latency


# Initialize the model
model = Model("PlacementPlan")

# Decision variables
x = model.addVars(sites, users, vtype=GRB.BINARY, name="x")  # Assignment of users to sites
y = model.addVars(sites, vtype=GRB.BINARY, name="y")  # Site is open
u = model.addVars(sites, vtype=GRB.INTEGER, name="u")  # Number of users per site


instance_per_site = model.addVars(sites, vtype=GRB.CONTINUOUS,name="instance_per_site") # Number of instances per site
utilization = model.addVars(sites, vtype=GRB.CONTINUOUS, name="utilization") # Utilization per site
response_time = model.addVars(sites, vtype=GRB.CONTINUOUS, name="response_times") # Response time per site
queuing_time = model.addVars(sites, vtype=GRB.CONTINUOUS, name="queuing_time") # Queuing time per site
rtt = model.addVars(sites, users, lb=-M, vtype=GRB.CONTINUOUS, name="rtt") # Round trip time


service_time = 1 / service_rate # Service time per instance

###################
if opti_pref == 1:
# # Objective: Minimize the number of open sites
    model.setObjective(y.sum(), GRB.MINIMIZE)
elif opti_pref == 2:
# Objective: Minimize the number of instances
    model.setObjective(instance_per_site.sum(), GRB.MINIMIZE)
elif opti_pref == 3:
# Objective: Minimize the latency
    model.setObjective(rtt.sum(), GRB.MINIMIZE)
else:
    print("No optimization preference selected.")
    sys.exit(1)
##################



###################
  # Constraints #
###################

# Each user is assigned to exactly one site
model.addConstrs((x.sum('*', j) == 1 for j in users), name="assignToOne")

# Users can only be linked to the open sites
model.addConstrs((x[i, j] <= y[i] for i in sites for j in users), name="linkToOpen")

# Capacity constraints: number of instance shouldn't be bigger than the number of slots
model.addConstrs((instance_per_site[i] * y[i] <= capacities[i] for i in sites), name="capacity")
# If a site is open, it should have at least one instance deployed
model.addConstrs((instance_per_site[i] >= y[i] for i in sites), name="atLeastOneInstanceIfOpen")

# Number of users per site
model.addConstrs((u[i] == x.sum(i, '*') for i in sites), name="numUsers")

# Enforce utilization
model.addConstrs((utilization[i] * service_rate * instance_per_site[i] == u[i] * request_rate for i in sites), name="utilization")

# Enforce queuing time
model.addConstrs((queuing_time[i] * (1 - utilization[i]) * service_rate * 2 == utilization[i] for i in sites), name="queuing_time")

# Enforce response time
model.addConstrs((response_time[i] == queuing_time[i] + service_time for i in sites), name="response_time")
# model.addConstrs((sum(i, '*') == 1 for i in sites), name="oneResponseTime")
# model.addConstrs((y[i] == sum([i,n] for n in response_time) for i in sites), name="matchResponseTime")

# Enforce response time only for opening sites
# model.addConstrs((response_time[i] == queuing_time[i] + service_time - M * (1 - y[i]) for i in sites), name="response_time_open")

# Enforce RTT
for j in users:
    for i in sites:
        model.addConstrs(
            (rtt[i, j] == latency_node_user[i, j] + response_time[i] - M * (1 - x[i, j]) for i in sites for j in users),
                          name=f"rtt_{i}_{j}"
                          )


# Enforce SLO
for j in users:
    for i in sites:
        model.addConstr(
            rtt[i, j] <= slo, name="slo"
            )


# Optimize the model
model.optimize()

# Print the optimal solution
if model.status == GRB.OPTIMAL:
    print("Optimal solution found:")
    for i in sites:
        if y[i].X > 0.5:
            print(f"Node {i} is open with {math.ceil(instance_per_site[i].X)} instances deployed and {u[i].X} users assigned.")
            for j in users:
                if x[i,j].X > 0.5:
                    print(f"User {j} for site: {i}, with RTT {rtt[i,j].X}.")
        else:
            print(f"Node {i} is closed.")
else:
    print("No optimal solution found.")


# Create the placement plan
placement_plan = {
    "node_replicas": {},
    "site_user_map": {},
    "deployment_name": deployment_name
}

for i in sites:
        if y[i].X > 0.5:  # If instance n is at site i
            placement_plan["node_replicas"][i] = math.ceil(instance_per_site[i].X)

for i in sites:
    for j in users:
        if x[i, j].X > 0.5:  # If user j is assigned to site i
            if i not in placement_plan["site_user_map"]:
                placement_plan["site_user_map"][i] = []
            placement_plan["site_user_map"][i].append(j)

now = datetime.now().strftime("%Y%m%d%H%M%S")

# Write the placement plan to a JSON file
with open(f"../data/placement/placement_plan_{now}.json", "w") as f:
    json.dump(placement_plan, f)


