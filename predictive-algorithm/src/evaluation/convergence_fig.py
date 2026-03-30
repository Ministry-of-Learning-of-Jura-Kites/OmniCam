import json

from matplotlib import pyplot as plt
import numpy as np
import seaborn as sns

with open("optimization_results_v1.json", "r") as f:
    import_data = json.load(f)

op_max_g = import_data["metadata"]["op_config"]["G"]
ref_max_g = import_data["metadata"]["ref_config"]["G"]


def pad_history_to_matrix(histories, max_g):
    """Safely pads varying length lists to max_g and returns a numpy matrix."""
    padded = []
    for h in histories:
        # Convert to list if it's not already
        h_list = list(h)

        # Pad with the last known value if the run terminated early
        if len(h_list) < max_g:
            last_val = h_list[-1]
            h_list.extend([last_val] * (max_g - len(h_list)))
        else:
            # Truncate if it somehow exceeded max_g
            h_list = h_list[:max_g]

        padded.append(h_list)
    return np.array(padded)


# 2. Extract History Lists
# We convert them to numpy arrays immediately for easier manipulation
op_fitness_histories = pad_history_to_matrix(
    import_data["op_fitness_histories"], op_max_g
)
ref_fitness_histories = pad_history_to_matrix(
    import_data["ref_fitness_histories"], ref_max_g
)

# 3. Extract Configuration/Max Generations
op_max_g = import_data["metadata"]["op_config"]["G"]
ref_max_g = import_data["metadata"]["ref_config"]["G"]

# 4. Optional: Re-calculate the global best for normalization
# This finds the absolute minimum cost across all reference trials
print("op max", np.max(op_fitness_histories))
global_best_fitness = np.max(ref_fitness_histories)

print(f"Successfully loaded {len(op_fitness_histories)} Operational trials.")
print(f"Successfully loaded {len(ref_fitness_histories)} Reference trials.")
print(f"Op Max G: {op_max_g} | Ref Max G: {ref_max_g}")
print(f"Global Minimum Cost Found: {global_best_fitness:.6f}")


def cost_to_fitness(cost_value):
    return 1 / cost_value


# --- Processing for Plotting ---
def process_to_fitness_matrix(histories, max_g):
    padded = []
    for h in histories:
        fitness_h = [c for c in h]
        if len(fitness_h) < max_g:
            p = np.pad(fitness_h, (0, max_g - len(fitness_h)), "edge")
        else:
            p = fitness_h[:max_g]
        padded.append(p)
    return np.array(padded)


op_matrix = process_to_fitness_matrix(op_fitness_histories, op_max_g)
ref_matrix = process_to_fitness_matrix(ref_fitness_histories, ref_max_g)

op_mean = np.mean(op_matrix, axis=0)

op_std = np.std(op_matrix, axis=0)

op_gens = np.arange(1, op_max_g + 1)


ref_mean = np.mean(ref_matrix, axis=0)

ref_std = np.std(ref_matrix, axis=0)

ref_gens = np.arange(1, ref_max_g + 1)

print("Final op mean", op_mean[-1])
print("Final ref mean", ref_mean[-1])
print("Eta", op_mean[-1] / ref_mean[-1] * 100)

avg_op_time = np.mean(import_data["summary"][0]["avg_op_time"])
avg_ref_time = np.mean(import_data["summary"][0]["avg_ref_time"])
speedup = avg_ref_time / avg_op_time

print(f"Mean Operational Time: {avg_op_time:.2f}s")
print(f"Mean Reference Time:   {avg_ref_time:.2f}s")
print(f"Average Speedup:        {speedup:.1f}x")

plt.figure(figsize=(10, 6))

# Academic theme

sns.set_theme(style="whitegrid")


# Define distinct colors for clarity

color_op = "#E69F00"  # High-visibility orange

color_ref = "#0072B2"  # Academic blue


# --- Plot Reference (Ref) ---

# Shaded variance (std deviation)

plt.fill_between(
    ref_gens, ref_mean - ref_std, ref_mean + ref_std, color=color_ref, alpha=0.15
)

# Mean line

plt.plot(
    ref_gens,
    ref_mean,
    color=color_ref,
    linewidth=2.5,
    label="Ref ($G_{max}=5000, N_p=200$)",
)

plt.fill_between(
    op_gens, op_mean - op_std, op_mean + op_std, color=color_op, alpha=0.3
)  # Slightly higher alpha for visibility

# Mean line

plt.plot(
    op_gens,
    op_mean,
    color=color_op,
    linewidth=2.5,
    linestyle="--",  # Dotted line to further distinguish
    label="Op ($G_{max}=500, N_p=50$)",
)


# --- Plot Final Optimality Point ---

# Explicitly mark the operational fitness achieved (f_op_avg)

final_op_fitness = op_mean[-1]

plt.scatter(
    op_max_g, final_op_fitness, color=color_op, s=80, edgecolors="black", zorder=5
)

plt.text(
    op_max_g * 1.1,
    final_op_fitness,
    f"{final_op_fitness:.2f}",
    color=color_op,
    fontweight="bold",
)


# Labels and Academic Tuning

plt.xscale("log")  # Use logarithmic scale for x-axis

plt.xlabel("Generation (Log Scale)", fontsize=12, fontweight="bold")

plt.ylabel("Fitness", fontsize=12, fontweight="bold")

plt.title("Convergence Efficiency Comparison", fontsize=14, fontweight="bold")

plt.grid(True, which="both", ls="-", alpha=0.5)


# Place legend in an empty area, typically lower right

plt.legend(loc="lower right", fontsize=10, frameon=True)


# Define axis limits slightly padded

plt.xlim(1, ref_max_g)


# Set tick parameters

plt.tick_params(axis="both", which="major", labelsize=10)


# Optional: Tight layout before saving

plt.tight_layout()

plt.savefig("convergence.pdf")

plt.show()
