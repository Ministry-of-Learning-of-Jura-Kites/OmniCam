import os
import sys

# This adds the 'messages' directory to the path
# so 'import camera_pb2' works for the generated files.
sys.path.append(os.path.dirname(__file__))
