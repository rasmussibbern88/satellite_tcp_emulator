import json
import jmespath

def parse_logfile(path: str):
    f = open(path,"r")
    lines = f.readlines()
    logs = { "log": []}
    for line in lines:
        data = json.loads(line)
        logs["log"].append(data)
    return logs



# logs_reno = parse_logfile("../logs_reno.json")
# logs_cubic = parse_logfile("../logs_cubic.json")
# logs_bbr = parse_logfile("../logs_bbr.json")
# logs_reno = parse_logfile("../emulatorJan_Reno.json")
# logs_cubic = parse_logfile("../emulatorJan_Cubic.json")
# logs_bbr = parse_logfile("../emulatorJan_BBR.json")
logs_reno = parse_logfile("../israel_reno.json")
logs_cubic = parse_logfile("../israel_cubic.json")

for fname, logs in [("cubic", logs_cubic), ("reno", logs_reno)]:
# for fname, logs in [("cubic", logs_cubic), ("reno", logs_reno), ("bbr", logs_bbr)]:
    
    #? Getting the time in the emulator when new paths are calculated 
    earthTime = jmespath.search("log[?message=='earthTime'].[index, earthTime, time]", logs)

    #? Getting the time a new path is being used. 
    newPath = jmespath.search("log[?message=='new Path'].[index, path, containers]", logs)

    debugPath = jmespath.search("log[?message=='debug path'].[path, time]", logs)

    path_change = jmespath.search("log[?message=='path change'].[pathDistance, nextPathDistance, time]", logs_reno)
    
    with open(f"paths_{fname}.json", 'w') as f:
        f.write(json.dumps(newPath))

    with open(f"realtime_{fname}.json", 'w') as f:
        f.write(json.dumps(earthTime))

    with open(f"debugPath_{fname}.json", 'w') as f:
        f.write(json.dumps(debugPath))

    with open(f"path_change_{fname}.json", 'w') as f:
        f.write(json.dumps(path_change))

#? Getting the ????
path_distance = jmespath.search("log[?message=='new path'].[path_distance, time]", logs_reno) # This is the first path

