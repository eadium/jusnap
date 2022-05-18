from setuptools import setup

setup(
    name="extkern",
    version="0.0.1",
    description="Kernel manager for connecting to persistent IPython kernel started outside of Jupyter",
    url="https://github.com/eadium/jusnap/tree/master/python_modules/jupyter/extkern",
    author="eadium",
    author_email="eadium@rasseki.org",
    license="MIT",
    packages=["extkern"],
    requires=["notebook"],
    zip_safe=False,
)
