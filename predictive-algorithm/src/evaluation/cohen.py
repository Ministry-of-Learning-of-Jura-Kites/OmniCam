import json

import numpy as np


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


def calculate_cohens_d_paired(group1, group2):
    """Calculates Cohen's d for paired samples."""
    diff = np.array(group1) - np.array(group2)
    return np.mean(diff) / np.std(diff, ddof=1)


# Extract the relevant data slices
de_results = export_de["results_gens"][:10]
pso_results = export_pso["results_gens"]
ga_results = export_ga["results_gens"]

# 1. DE vs PSO Effect Size
d_pso = calculate_cohens_d_paired(de_results, pso_results)

# 2. DE vs GA Effect Size
d_ga = calculate_cohens_d_paired(de_results, ga_results)

print(f"Cohen's d (DE vs PSO): {d_pso:.4f}")
print(f"Cohen's d (DE vs GA): {d_ga:.4f}")


# Interpretation based on Cohen (1988)
def interpret_cohen(d):
    d = abs(d)
    if d < 0.2:
        return "Negligible"
    if d < 0.5:
        return "Small"
    if d < 0.8:
        return "Medium"
    return "Large"


print(f"Effect size for PSO is {interpret_cohen(d_pso)}")
print(f"Effect size for GA is {interpret_cohen(d_ga)}")
