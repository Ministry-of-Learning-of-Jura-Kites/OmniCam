import json
import matplotlib.pyplot as plt
import scipy.stats as stats
import seaborn as sns

# File list
graphs = ["case a.json", "case b.json", "case c.json"]
names = ["Scenario A", "Scenario B", "Scenario C"]


def generate_distribution_report(
    file_list, names, output_name="scenarios_normality_horizontal.pdf"
):
    num_cases = len(file_list)

    # Grid: 2 rows (Distributions, Q-Q Plots), num_cases columns (Scenarios)
    # Adjusted figsize to be wider for more scenarios
    fig, axes = plt.subplots(2, num_cases, figsize=(5 * num_cases, 9), squeeze=False)

    for i, filename in enumerate(file_list):
        try:
            with open(filename, "r") as f:
                data = json.load(f)
                costs = data["costs"]
        except FileNotFoundError:
            print(f"File {filename} not found.")
            continue

        # --- Row 1: Cost Distribution (Histogram) ---
        ax_hist = axes[0, i]
        sns.histplot(costs, kde=True, ax=ax_hist, color="steelblue")

        # Scenario Name as Column Title
        ax_hist.set_title(names[i], fontsize=16, pad=15, fontweight="bold")
        ax_hist.set_xlabel("Final Cost", fontsize=11)

        # Label the Row (Only on the leftmost plot)
        if i == 0:
            ax_hist.set_ylabel(
                "Cost Distribution\n\nFrequency", fontweight="bold", fontsize=12
            )
        else:
            ax_hist.set_ylabel("Frequency")

        # --- Row 2: Normal Q-Q Plot ---
        ax_qq = axes[1, i]
        stats.probplot(costs, dist="norm", plot=ax_qq)

        # Styling the plot points
        ax_qq.get_lines()[0].set_markerfacecolor("grey")
        ax_qq.get_lines()[0].set_alpha(0.5)
        ax_qq.get_lines()[0].set_markersize(5)

        # Fixing labels and removing the default title
        ax_qq.set_title("")
        ax_qq.set_xlabel("Theoretical Quantiles", fontsize=11)

        # Label the Row (Only on the leftmost plot)
        if i == 0:
            ax_qq.set_ylabel(
                "Normal Q-Q Plot\n\nOrdered Values", fontweight="bold", fontsize=12
            )
        else:
            ax_qq.set_ylabel("Ordered Values")

    plt.tight_layout()
    plt.savefig(output_name, bbox_inches="tight")
    print(f"Successfully saved horizontal analysis to {output_name}")
    plt.show()


if __name__ == "__main__":
    generate_distribution_report(graphs, names)
