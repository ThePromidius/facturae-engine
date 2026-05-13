import os
import sys
from datetime import datetime

# -- Project information -----------------------------------------------------
project = 'FacturaE Sidecar'
copyright = f'{datetime.now().year}, Victor'
author = 'Victor'
release = '0.1.0'

# -- General configuration ---------------------------------------------------
extensions = [
    'myst_parser',           # For Markdown support
    'sphinx.ext.autodoc',    # For documentation from docstrings
    'sphinx.ext.viewcode',   # To add links to highlighted source code
    'sphinx.ext.todo',       # To support todo items
    'sphinx.ext.extlinks',   # For external links to source repo
]

# -- Extlinks configuration --------------------------------------------------
# This allows us to link to source code on GitHub easily
extlinks = {
    'src': ('https://github.com/ThePromidius/facturae-engine/blob/main/%s', '%s'),
}

# Add support for both .rst and .md
source_suffix = {
    '.rst': 'restructuredtext',
    '.md': 'markdown',
}

templates_path = ['_templates']
exclude_patterns = ['_build', 'Thumbs.db', '.DS_Store', '.venv']

# -- Options for HTML output -------------------------------------------------
html_theme = 'sphinx_rtd_theme'
html_static_path = ['_static']

# -- MyST Parser configuration -----------------------------------------------
myst_enable_extensions = [
    "amsmath",
    "colon_fence",
    "deflist",
    "dollarmath",
    "fieldlist",
    "html_admonition",
    "html_image",
    "linkify",
    "replacements",
    "smartquotes",
    "substitution",
    "tasklist",
]
