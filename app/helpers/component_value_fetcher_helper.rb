# frozen_string_literal: true

module ComponentValueFetcherHelper
  def component_value_fetcher_tag(controller_name, url)
    tag.div(
      data: {
        controller: "component-value-fetcher--#{controller_name}",
        "component-value-fetcher--#{controller_name}-url-value": url,
        "component-value-fetcher--#{controller_name}-event-name-value": "component-value-fetcher:#{controller_name}:fetched"
      }
    )
  end
end
