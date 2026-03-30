from typing import Union
from pyvistaqt import BackgroundPlotter
import numpy as np
from state import State
from cost_functions import total_cost
from . import CartesianSerialize


def optimize_ga(
    initial_state: State,
    pl: Union[BackgroundPlotter, None],
    seed: int = 2000,
    iterations=500,
    pop_size=50,
):
    cartesian = CartesianSerialize(seed)
    num_cams = len(initial_state.cameras)
    initial_vec = cartesian.state_to_vector(initial_state)
    bounds = cartesian.init_bounds(initial_state)

    # 1. Initialize Population
    # We use your existing logic to create the initial pool of candidates
    population = cartesian.init_pop(pop_size, initial_vec, bounds, num_cams)
    population += np.random.uniform(-0.5, 0.5, population.shape)

    # GA Hyperparameters
    mutation_rate = 0.1
    crossover_rate = 0.8
    elite_size = 2  # Keep the best performers unchanged

    # Initial Fitness Evaluation
    costs = np.array(
        [total_cost(cartesian.vector_to_state(p, initial_state)) for p in population]
    )

    best_idx = np.argmin(costs)
    g_best = population[best_idx].copy()
    g_best_cost = costs[best_idx]

    # --- Tracking Variables ---
    generations = 0
    patience = 50
    min_delta = 1e-2
    no_improvement_counter = 0
    # --------------------------

    for iteration in range(iterations):
        generations += 1
        prev_g_best_cost = g_best_cost

        # 2. Selection (Tournament Selection)
        new_population = []

        # Elitism: Carry over the best
        elite_indices = np.argsort(costs)[:elite_size]
        for idx in elite_indices:
            new_population.append(population[idx].copy())

        while len(new_population) < pop_size:
            # Tournament Selection
            indices = np.random.choice(pop_size, 3)
            parent1 = population[indices[np.argmin(costs[indices])]]

            indices = np.random.choice(pop_size, 3)
            parent2 = population[indices[np.argmin(costs[indices])]]

            # 3. Crossover (Heuristic/Arithmetic Crossover)
            if np.random.rand() < crossover_rate:
                alpha = np.random.rand(*parent1.shape)
                child = alpha * parent1 + (1 - alpha) * parent2
            else:
                child = parent1.copy()

            # 4. Mutation
            if np.random.rand() < mutation_rate:
                mutation_noise = np.random.normal(0, 0.2, size=child.shape)
                child += mutation_noise

            # Clip to bounds if necessary
            # child = np.clip(child, bounds[:, 0], bounds[:, 1])

            new_population.append(child)

        population = np.array(new_population)

        # 5. Evaluate New Population
        costs = np.array(
            [
                total_cost(cartesian.vector_to_state(p, initial_state))
                for p in population
            ]
        )

        # Update Global Best
        current_best_idx = np.argmin(costs)
        if costs[current_best_idx] < g_best_cost:
            g_best_cost = costs[current_best_idx]
            g_best = population[current_best_idx].copy()

        # --- Early Stopping Logic ---
        if (prev_g_best_cost - g_best_cost) < min_delta:
            no_improvement_counter += 1
        else:
            no_improvement_counter = 0

        if no_improvement_counter >= patience:
            print(
                f"Early stopping at iteration {iteration} due to lack of improvement."
            )
            break

    return cartesian.vector_to_state(g_best, initial_state), generations
