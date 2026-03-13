import cv2
import matplotlib.pyplot as plt


def plot_pixel_difference(image_path1, image_path2):
    # Load images in grayscale
    # Using cv2.IMREAD_GRAYSCALE ensures we are comparing intensity
    img1 = cv2.imread(image_path1, cv2.IMREAD_GRAYSCALE)
    img2 = cv2.imread(image_path2, cv2.IMREAD_GRAYSCALE)

    if img1 is None or img2 is None:
        print("Error: Could not load one or both images. Please check the file paths.")
        return

    # Ensure dimensions match
    if img1.shape != img2.shape:
        print(
            f"Error: Image dimensions must match. Image 1 is {img1.shape} and Image 2 is {img2.shape}."
        )
        return

    # Calculate absolute difference
    # Let $I_1$ and $I_2$ be the intensity values of the two images.
    # We calculate the absolute difference $D = |I_1 - I_2|$.
    diff = cv2.absdiff(img1, img2)

    # Plot the histogram
    plt.figure(figsize=(10, 6))
    plt.hist(
        diff.ravel(),
        bins=256,
        range=[0, 10],
        color="blue",
        alpha=0.7,
        edgecolor="black",
    )
    plt.title("Histogram of Pixel Intensity Differences")
    plt.xlabel("Intensity Difference Value")
    plt.ylabel("Frequency")
    plt.grid(axis="y", alpha=0.5)

    # Ensure labels are not truncated
    plt.tight_layout()

    # Save the plot
    plt.show()
    print("Histogram saved as 'pixel_difference_histogram.png'.")


plot_pixel_difference(
    "/home/frook/Downloads/from_algo_service.png",
    "/home/frook/Downloads/from_frontend.png",
)
