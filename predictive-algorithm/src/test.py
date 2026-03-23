import re
import sys


def check_latex_citation_order(file_path):
    with open(file_path, "r", encoding="utf-8") as f:
        content = f.read()

    # 1. Find the order of appearances of \cite{...}
    # This handles multiple citations like \cite{ref1, ref2}
    citations = re.findall(r"\\cite\{([^{}]+)\}", content)
    appearance_order = []
    for c in citations:
        keys = [k.strip() for k in c.split(",")]
        for key in keys:
            print(keys)
            if key not in appearance_order:
                appearance_order.append(key)

    # 2. Find the order of \bibitem{...} in the bibliography
    bib_order = re.findall(r"\\bibitem\{([^{}]+)\}", content)

    if not bib_order:
        print("Error: No \\bibitem entries found. Are they in a separate .bib file?")
        return

    # 3. Compare
    print(f"Found {len(appearance_order)} unique citations in text.")
    print(f"Found {len(bib_order)} entries in bibliography.\n")

    errors = 0
    for i, (cite_key, bib_key) in enumerate(zip(appearance_order, bib_order)):
        if cite_key != bib_key:
            print(f"[X] MISMATCH at index {i+1}:")
            print(f"    Expected (from text): {cite_key}")
            print(f"    Found (in bib block): {bib_key}")
            errors += 1

    if errors == 0 and len(appearance_order) == len(bib_order):
        print("Success: Reference order is perfectly aligned with text appearance!")
    elif len(appearance_order) != len(bib_order):
        print(
            f"\nWarning: Count mismatch! Text has {len(appearance_order)}, Bib has {len(bib_order)}."
        )


# Run the script on your file
check_latex_citation_order("main.tex")
