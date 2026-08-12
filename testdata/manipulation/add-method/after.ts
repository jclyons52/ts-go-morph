export class Counter {
  count: number = 0;

  increment(by: number): void {
    this.count += by;
  }
}
