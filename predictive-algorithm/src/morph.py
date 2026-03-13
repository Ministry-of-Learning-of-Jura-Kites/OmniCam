import cv2


img = cv2.imread("/home/frook/Downloads/depth_map_save.png", cv2.IMREAD_GRAYSCALE)

kernel = cv2.getStructuringElement(cv2.MORPH_RECT, (2, 2))

# Apply Opening to remove speckle noise
noise_reduced = cv2.morphologyEx(img, cv2.MORPH_OPEN, kernel)

cv2.imwrite("/home/frook/Downloads/depth_map_save_morphed.png", noise_reduced)
