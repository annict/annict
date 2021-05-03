# frozen_string_literal: true

module ComponentValueFetcherHelper
  def component_value_fetcher_tag(component_name, url)
    tag.div(
      data: {
        controller: "component-value-fetcher",
        component_value_fetcher_url_value: url,
        component_value_fetcher_event_name_value: "component-value-fetcher:#{component_name}:fetched"
      }
    )
  end
end
