import re
from collections import OrderedDict


def analyze_latex_citations(tex_file_path):
    # Regex patterns
    # Matches a % and everything following it on that line
    comment_pattern = re.compile(r"%.*")
    cite_pattern = re.compile(r"\\cite\{([^{}]+)\}")
    bib_pattern = re.compile(r"\\bibitem\{([^{}]+)\}|@\w+\{([^,]+),")

    cited_keys = []
    defined_keys = []

    try:
        with open(tex_file_path, "r", encoding="utf-8") as f:
            # Read line by line to handle comments more cleanly
            lines = f.readlines()

            # Remove comments from every line and join back into a single string
            clean_content = ""
            for line in lines:
                # This removes the % and everything after it
                clean_line = re.sub(comment_pattern, "", line)
                clean_content += clean_line + "\n"

            # 1. Find all citations in order of appearance
            raw_cites = cite_pattern.findall(clean_content)
            for group in raw_cites:
                # Split by comma for multiple keys: \cite{key1, key2}
                for key in group.split(","):
                    cited_keys.append(key.strip())

            # 2. Find all defined bibliography keys
            raw_defs = bib_pattern.findall(clean_content)
            for item in raw_defs:
                # Filter out empty matches from the regex groups
                key = item[0] if item[0] else item[1]
                defined_keys.append(key.strip())

    except FileNotFoundError:
        print("Error: File not found.")
        return

    # Process Data
    unique_cited = list(OrderedDict.fromkeys(cited_keys))
    unused_keys = [k for k in defined_keys if k not in cited_keys]
    undefined_citations = [k for k in unique_cited if k not in defined_keys]

    # Report Generation
    print("--- LaTeX Citation Analysis (Comments Ignored) ---")
    print(f"Total Citations Found: {len(cited_keys)}")
    print(f"Unique Keys Used: {len(unique_cited)}")

    print("\n[1] Order of Appearance:")
    for i, key in enumerate(unique_cited, 1):
        print(f"{i}. {key}")

    if unused_keys:
        print("\n[2] Unused References (Defined but never cited):")
        for key in unused_keys:
            print(f" - {key}")

    if undefined_citations:
        print("\n[3] Warning: Broken Citations (Cited but not defined):")
        for key in undefined_citations:
            print(f" !! {key}")

    print("\n" + "=" * 30)
    print("Mitigation & Reordering Plan")
    print("=" * 30)

    if unused_keys:
        print(
            f"ADVICE: Remove the {len(unused_keys)} unused entries to reduce clutter."
        )

    print("To sync your bibliography with the order of appearance:")
    for i, key in enumerate(unique_cited, 1):
        print(f" Step {i}: Place '{key}' in position {i} of your .bib/bibitem list.")


# Usage
analyze_latex_citations("main.tex")
