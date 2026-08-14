import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { SearchableCatalog } from "./SearchableCatalog";

it("filters names without changing source selection", async () => {
  const user = userEvent.setup();
  render(
    <SearchableCatalog
      label="Buscar"
      items={[
        { id: 1, name: "Sentadilla" },
        { id: 2, name: "Plancha" },
      ]}
    >
      {(item) => <button>{item.name}</button>}
    </SearchableCatalog>,
  );
  await user.type(screen.getByRole("searchbox"), "plan");
  expect(screen.getByRole("button", { name: "Plancha" })).toBeInTheDocument();
  expect(
    screen.queryByRole("button", { name: "Sentadilla" }),
  ).not.toBeInTheDocument();
});
