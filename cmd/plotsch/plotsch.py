#!/usr/bin/env python3

'''
Plot a schematic to JSON/SVG with plotgitsch

SPDX-License-Identifier: MIT

Copyright (c) 2021 Fully Automated OÜ
'''

import argparse
import tempfile
import shutil
import os
import subprocess
import sys
import json


def parse_cli_args():
    parser = argparse.ArgumentParser(description='Plot Schematic')
    parser.add_argument('-i', "--input_folder",
                        type=str, help="Input folder")
    parser.add_argument('-a',
                        type=str, help="Revision A to compare to")
    parser.add_argument('-b',
                        type=str, help="Revision B to compare to")
    args = parser.parse_args()
    return args


if __name__ == "__main__":
    args = parse_cli_args()

    repo = args.input_folder

    os.chdir(repo)
    command = subprocess.run(
        ['plotgitsch', '-k', '-i', 'echo', args.a, args.b], capture_output=True)

    if command.returncode != 0:
        sys.exit(command.returncode)

    lines = command.stdout.splitlines()

    schematics = dict()
    clean = False

    if shutil.which("svgcleaner"):
        clean = True

    for line in lines:
        file = line.decode("utf-8")
        if file.startswith('internal diff'):
            continue
        # svgcleaner is incredibly fast and reduces the size of the svgs we have to serve
        if clean:
            command = subprocess.run(
                ['svgcleaner', file, file], capture_output=True)
            if command.returncode != 0:
                sys.exit(command.returncode)

        # read the svg and store it in our dict
        with open(file) as f:
            schematics[os.path.basename(file)] = f.read()
        os.remove(file)

    print(json.dumps({
        "schematics": schematics
    }))
