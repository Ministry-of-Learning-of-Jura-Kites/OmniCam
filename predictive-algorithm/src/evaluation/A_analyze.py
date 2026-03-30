import json
import math

import numpy as np

super = None
with open("case b pso.json", "r") as json_file:
    export = json.load(json_file)  #

# with open("case a super.json", "r") as json_file:
#     super = json.load(json_file)  #

times = export["times"]
costs = export["costs"]
results_gens = export["results_gens"]
seeds = export["seeds"]
avg_time = sum(times) / len(times)
avg_cost = sum(costs) / len(costs)
med_cost = np.median(costs)

gen_mean = np.mean(results_gens)
gen_std = np.std(results_gens)

n = len(costs)
mean_cost = np.mean(costs)
std_err = np.std(costs, ddof=1) / math.sqrt(n)  # ddof=1 for sample std dev

# -------------
if super is not None:
    super_cost = super["costs"]
    avg_super_cost = np.mean(super_cost)
    epsilon = 1e-10

    # 2. Calculate the average cost of your current case
    avg_current_cost = np.mean(costs)

    # 3. Quality Ratio (QR)
    # Closer to 1.0 means it is performing as well as the 'Super' case
    quality_ratio = (min(super_cost) + epsilon) / (min(costs) + epsilon)

    avg_super_times = sum(super["times"]) / len(super["times"])

    # 4. Computational Budget Ratio (BR)
    # How much 'cheaper' was this run?
    # (Total generations * Population size) or just Time
    budget_ratio = avg_time / avg_super_times

    # 5. η (Efficiency Index)
    # High η means you got a high-quality solution for a very low budget.
    eta = quality_ratio / budget_ratio

    eta_pct = eta * 100

# ---------------

ci_lower = mean_cost - (1.96 * std_err)
ci_upper = mean_cost + (1.96 * std_err)

SUCCESS_THRESHOLD = 1

success = [1 if c <= SUCCESS_THRESHOLD else 0 for c in costs]
success_pct = (sum(success) / len(seeds)) * 100

efficiency = avg_time / (success_pct / 100) * 100 if success_pct > 0 else float("inf")

print("\n" + "=" * 30)
if super is not None:
    print(f"Quality Ratio vs Super: {quality_ratio:.4f}")
    print(f"Effiency %: {quality_ratio*100:.4f}")
    print(f"Efficiency Percent (η):   {eta_pct:.4f}")
    print(f"Avg super time:   {np.mean(super["times"]):.4f}")
print(f"ci_lower:   {ci_lower}")
print(f"ci_upper:   {ci_upper}")
print(f"Success %:   {success_pct}")
print(f"Converged %:   {success_pct}")
print(f"RESULTS FOR {len(seeds)} RUNS")
print(f"Generations:  {gen_mean} ± {gen_std}")
print(f"Average Elapsed Time: {avg_time} seconds")
print(f"Average Total Cost:   {avg_cost}")
print(f"Best Cost:   {min(costs)}")
print("=" * 30)
