export const existing = 1;

export class Person {
  name: string;

  constructor(name: string) {
    this.name = name;
  }

  greet(greeting: string): string {
    return greeting + " " + this.name;
  }
}
