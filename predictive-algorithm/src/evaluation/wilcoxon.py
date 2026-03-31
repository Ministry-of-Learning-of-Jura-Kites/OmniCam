import json
from scipy.stats import wilcoxon

with open("case b pso.json", "r") as json_file:
    export_pso = json.load(json_file)  #

with open("case b.json", "r") as json_file:
    export_de = json.load(json_file)  #

with open("case b ga.json", "r") as json_file:
    export_ga = json.load(json_file)  #

times_pso = export_pso["times"]
costs_pso = export_pso["costs"]
results_gens_pso = export_pso["results_gens"]
seeds_pso = export_pso["seeds"]

# 1. DE vs PSO
stat_pso, p_val_pso = wilcoxon(
    export_de["results_gens"][:10], export_pso["results_gens"]
)

# 2. DE vs GA
stat_ga, p_val_ga = wilcoxon(export_de["results_gens"][:10], export_ga["results_gens"])

print(f"DE vs PSO p-value: {p_val_pso:.4f}")
print(f"DE vs GA p-value: {p_val_ga:.4f}")

# Optional: Check significance at alpha = 0.05
for name, p in [("PSO", p_val_pso), ("GA", p_val_ga)]:
    status = "Significant" if p < 0.05 else "Not Significant"
    print(f"Difference between DE and {name} is {status}")
