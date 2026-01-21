// JavaScript Test File

// explain javascript to me

// convert to fat arrow style
function heavyComputation(a, b) {
    const factor = 2.5;
    let result = 0;

    // Perform loop
    for (let i = 0; i < 10; i++) {
        if (i % 2 === 0) {
            result += (a * b) * factor;
        } else {
            result -= i;
        }
    }

    return result;
}

const user = {
    name: "John Doe",
    age: 30,
    isActive: true
};

console.log("Starting computation...");
var output = heavyComputation(10, 20);
console.log(`Final Result: ${output}`);

class Processor {
    constructor(data) {
        this.data = data;
    }

    process() {
        return this.data.map(x => x * 2);
    }
}
