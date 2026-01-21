# A simple Ruby class demonstration

class Person
  attr_accessor :name, :age

  def initialize(name, age)
    @name = name
    @age = age
  end

  def introduce
    puts "Hi, I'm #{@name} and I am #{@age} years old."
  end

  def can_vote?
    @age >= 18
  end
end

# Module definition
module Greeter
  def self.say_hello(name)
    puts "Hello, #{name}!"
  end
end

# Main execution
if __FILE__ == $0
  alice = Person.new("Alice", 25)
  alice.introduce

  if alice.can_vote?
    puts "#{alice.name} can vote."
  else
    puts "#{alice.name} cannot vote."
  end

  Greeter.say_hello("Bob")

  # Array and block
  numbers = [1, 2, 3, 4, 5]
  squared = numbers.map { |n| n * n }
  puts "Squared numbers: #{squared.inspect}"

  # Hash
  config = {
    :env => "production",
    :retries => 3,
    :timeout => 500
  }
  
  puts "Environment: #{config[:env]}"
end
