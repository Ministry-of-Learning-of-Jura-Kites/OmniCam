import numpy as np

x = [114652, 557659, 1164042, 4656168]
acc = [
    0.000054,
    0.000089,
    0.000147,
    0.000207,
]
brute_force = [0.032709, 0.135381, 0.214512, 2.238618]


def calculate_empirical_complexity(n, t):
    # 1. Transform data to log-log scale
    log_n = np.log(n)
    log_t = np.log(t)

    # 2. Perform linear regression (polyfit of degree 1)
    # Returns [slope, intercept]
    slope, intercept = np.polyfit(log_n, log_t, 1)

    return slope, intercept


print("acc", calculate_empirical_complexity(x, acc))
print("brute_force", calculate_empirical_complexity(x, brute_force))
