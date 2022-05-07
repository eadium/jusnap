from setuptools import setup

setup(
    name="extkern",
    version="0.0.1",
    description="Kernel manager for connecting to persistent IPython kernel started outside of Jupyter",
    url="http://github.com/ebanner/extipy",
    author="eadium",
    author_email="eadium@rasseki.org",
    license="MIT",
    packages=["extkern"],
    requires=["notebook"],
    zip_safe=False,
)
