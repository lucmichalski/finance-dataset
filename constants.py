import json
from collections import namedtuple
from pprint import pprint


def convert(d):
    # dict -> namedtuple
    return namedtuple('GenericDict', d.keys())(**d)


with open('conf.json') as data_file:
    data = json.load(data_file)
    print('Configuration filename')
    pprint(data, indent=2)
    data = convert(data)
