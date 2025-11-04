export function safeGetAspectRatio(aspectWidth: number, aspectHeight: number) {
  if (aspectHeight == 0) {
    return 0;
  }
  return aspectWidth / aspectHeight;
}
