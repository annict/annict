# frozen_string_literal: true

module ReactHelper
  def react_component(name, options = {})
    tag.div(
      id: "ann-app",
      class: " d-flex flex-column",
      "data-component-name": name
    )
  end
end
