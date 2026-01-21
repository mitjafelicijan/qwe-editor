def factorial(n):
    # This is a comment explaining the function
    if n <= 1:
        return 1
    else:
        return n * factorial(n - 1)

class Calculator:
    def __init__(self):
        self.result = 0

    def add(self, a, b):
        """Docstring for add method"""
        return a + b

if __name__ == "__main__":
    print("Factorial of 5 is:", factorial(5))
    calc = Calculator()
    res = calc.add(10, 20)
    print(f"Result: {res}")
