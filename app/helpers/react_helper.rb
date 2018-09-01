# frozen_string_literal: true

module ReactHelper
  def react_component(name, options = {})
    tag.div(id: "ann-react-app", "data-component-name": name)
  end
end
