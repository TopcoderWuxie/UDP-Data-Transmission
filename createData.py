#coding: utf-8

import random
import string

lower = [x for x in string.lowercase]
upper = [x for x in string.uppercase]
results = []

def main():
	res = []
	ip = []
	for i in range(4):
		ip.append(str(random.randint(0, 256)))

	ip = ".".join(ip)
	res.append(ip)

	random.shuffle(lower)
	requested = "".join(lower)
	res.append(requested)

	random.shuffle(upper)
	traced = "".join(upper)
	res.append(traced)

	random.shuffle(lower)
	api = "".join(lower)
	res.append(api)

	mean = str(random.randint(0, 1000)) + "ms"
	res.append(mean)

	random.shuffle(upper)
	comment = "".join(upper)
	res.append(comment)
	
	res = "|".join(res)
	results.append(res)
	# print ip, requested, traced, api, mean, comment

if __name__ == "__main__":
	for i in range(10000):
		main()
	with open("text.txt", "wb") as f:
		f.write("\n".join(results))