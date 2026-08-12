export interface Empty {
  id: number;
}

export interface Named {
  name: string;

  rename(newName: string): void;
}
