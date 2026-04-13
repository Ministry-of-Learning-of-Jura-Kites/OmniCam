import json

import numpy as np
import matplotlib.pyplot as plt

super = None
with open("case a.json", "r") as json_file:
    export = json.load(json_file)  #

times = export["times"]
costs = export["costs"]


def analyze_sufficiency(data, n_iterations=10000):
    n = len(data)
    bootstrapped_means = []

    # 1. Perform Bootstrapping
    for _ in range(n_iterations):
        # Resample with replacement
        sample = np.random.choice(data, size=n, replace=True)
        bootstrapped_means.append(np.mean(sample))

    # 2. Calculate Non-Parametric Confidence Interval (95%)
    # This doesn't care about skewness!
    ci_low = np.percentile(bootstrapped_means, 2.5)
    ci_high = np.percentile(bootstrapped_means, 97.5)
    ci_width = ci_high - ci_low

    # 3. Calculate Stability (Relative Error)
    # If the width of our uncertainty is < 5-10% of the mean, we are usually safe.
    mean_val = np.mean(data)
    stability_ratio = (ci_width / mean_val) * 100

    print("\n" + "BOOTSTRAP SUFFICIENCY CHECK" + "\n" + "=" * 30)
    print(f"Number of Seeds: {n}")
    print(f"Bootstrap Mean:  {np.mean(bootstrapped_means):.4f}")
    print(f"95% Bootstrap CI: [{ci_low:.4f}, {ci_high:.4f}]")
    print(f"CI Width:        {ci_width:.4f}")
    print(f"Stability Ratio: {stability_ratio:.2f}% (Width as % of Mean)")

    if stability_ratio < 10:
        print("STATUS: SUFFICIENT. The mean is stable across seeds.")
    elif stability_ratio < 20:
        print("STATUS: MARGINAL. You might need 50-100 seeds for better precision.")
    else:
        print("STATUS: INSUFFICIENT. High skew/variance detected. Add more seeds.")
    print("=" * 30)

    # 4. Optional: Convergence Plot
    # Shows how the average "settles" as you add more seeds
    cumulative_avg = np.cumsum(data) / (np.arange(n) + 1)
    plt.figure(figsize=(10, 5))
    plt.plot(range(1, n + 1), cumulative_avg, marker="o", linestyle="-")
    plt.axhline(y=mean_val, color="r", linestyle="--", label="Final Mean")
    plt.title("Mean Cost Convergence across Seeds")
    plt.xlabel("Number of Seeds")
    plt.ylabel("Cumulative Average Cost")
    plt.grid(True)
    plt.show()


# Run the check
analyze_sufficiency(costs)
