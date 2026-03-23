import cv2
import numpy as np

# Load images in grayscale for simplicity (color also works but requires handling 3 channels)
img1 = cv2.imread(
    "/home/frook/Downloads/morphed_reverted_from_algo.png", cv2.IMREAD_GRAYSCALE
)
img2 = cv2.imread(
    "/home/frook/Downloads/reverted_from_frontend.png", cv2.IMREAD_GRAYSCALE
)

# Ensure images have the same dimensions before comparison
if img1.shape != img2.shape:
    print("Images must have the same dimensions")
    # Optional: Resize one image to match the other
    # img2 = cv2.resize(img2, (img1.shape[1], img1.shape[0]))

# Calculate the absolute difference between the two images
# cv2.absdiff handles normalization so no negative values are returned
diff = cv2.absdiff(img1, img2)

cv2.imwrite("/home/frook/Downloads/depth_map_save.png", diff)
