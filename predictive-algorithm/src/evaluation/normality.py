import json
import matplotlib.pyplot as plt
import scipy.stats as stats
import seaborn as sns

# File list
graphs = ["case a.json", "case b.json", "case c.json"]
names = ["Scenario A", "Scenario B", "Scenario C"]


def generate_distribution_report(
    file_list, names, output_name="scenarios_normality.pdf"
):
    num_cases = len(file_list)
    fig, axes = plt.subplots(num_cases, 2, figsize=(10, 3.5 * num_cases))

    if num_cases == 1:
        axes = axes.reshape(1, 2)

    # Set Column Headers (Only on the top row)
    # We use larger titles here to ensure they stand out as group headers
    axes[0, 0].set_title("Cost Distribution", fontsize=16, pad=25, fontweight="bold")
    axes[0, 1].set_title("Normal Q-Q Plot", fontsize=16, pad=25, fontweight="bold")

    for i, filename in enumerate(file_list):
        try:
            with open(filename, "r") as f:
                data = json.load(f)
                costs = data["costs"]
        except FileNotFoundError:
            print(f"File {filename} not found.")
            continue

        # --- Column 1: Histogram ---
        sns.histplot(costs, kde=True, ax=axes[i, 0], color="steelblue")
        axes[i, 0].set_ylabel(
            f"{names[i]}\n\nFrequency", fontweight="bold", fontsize=12
        )
        axes[i, 0].set_xlabel("Final Cost", fontsize=11)

        # --- Column 2: Q-Q Plot ---
        stats.probplot(costs, dist="norm", plot=axes[i, 1])

        # 1. Manually override the title probplot creates
        if i == 0:
            axes[i, 1].set_title(
                "Normal Q-Q Plot", fontsize=16, pad=25, fontweight="bold"
            )
        else:
            axes[i, 1].set_title("")  # Remove titles for subsequent rows

        # 2. Styling the plot points
        axes[i, 1].get_lines()[0].set_markerfacecolor("grey")
        axes[i, 1].get_lines()[0].set_alpha(0.5)
        axes[i, 1].get_lines()[0].set_markersize(5)

        # 3. Fixing labels that probplot overrides
        axes[i, 1].set_xlabel("Theoretical Quantiles", fontsize=11)
        axes[i, 1].set_ylabel("Ordered Values", fontsize=11)

        # --- Clean up Repetitive Titles ---
        if i > 0:
            axes[i, 0].set_title("")

    plt.tight_layout()
    plt.savefig(output_name, bbox_inches="tight")
    print(f"Successfully saved analysis to {output_name}")
    plt.show()


if __name__ == "__main__":
    generate_distribution_report(graphs, names)
