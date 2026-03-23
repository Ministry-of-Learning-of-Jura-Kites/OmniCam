import cv2
import numpy as np

img1 = cv2.imread("/home/frook/Downloads/reverted_from_algo.png", cv2.IMREAD_GRAYSCALE)
img2 = cv2.imread(
    "/home/frook/Downloads/reverted_from_frontend.png", cv2.IMREAD_GRAYSCALE
)
# mask1 = img1 > 0
# mask2 = img2 > 0

# # 3. Calculate Logical Operations
# intersection = np.logical_and(mask1, mask2)
# union = np.logical_or(mask1, mask2)

# # 4. Calculate Metrics
# # Intersection over Union (IoU) - The academic standard
# iou_score = np.sum(intersection) / np.sum(union)

# # Pixel Accuracy (How much of the total image matches)
# # This includes matching background pixels
# total_pixels = img1.shape[0] * img1.shape[1]
# matching_pixels = np.sum(mask1 == mask2)
# pixel_accuracy = (matching_pixels / total_pixels) * 100

diff_values = (img1 / 255).astype(float) - (img2 / 255).astype(float)
rmse = np.sqrt(np.mean(diff_values**2))
print(f"Depth Value RMSE: {rmse:.4f} units")

# print(f"Total Pixel Accuracy: {pixel_accuracy:.2f}%")
