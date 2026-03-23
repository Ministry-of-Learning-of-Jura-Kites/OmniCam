import matplotlib.pyplot as plt
import imageio.v3 as iio

# Load your three images
fe = iio.imread("/home/frook/Downloads/from_frontend.png")
algo = iio.imread("/home/frook/Downloads/from_algo_service.png")
absdiff = iio.imread("/home/frook/Downloads/depth_map_save_morphed.png")

# Create figure with 3 subplots
fig, axes = plt.subplots(1, 3, figsize=(15, 5))

# Plotting
titles = ["(a) Frontend Viewport", "(b) Algorithmic Result", "(c) Residual Map"]
images = [fe, algo, absdiff]

for i, ax in enumerate(axes):
    ax.imshow(images[i], cmap="gray", vmin=0, vmax=255)
    ax.set_title(titles[i], fontsize=12)
    ax.axis("off")  # Hide axis ticks for clean view

plt.tight_layout()
plt.savefig("synchronization_figure.pdf", dpi=300)
