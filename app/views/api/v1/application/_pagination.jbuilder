# frozen_string_literal: true

json.total_count collection.total_count
json.next_page collection.last_page? ? nil : (params.page.to_i + 1)
json.prev_page params.page.to_i > 1 ? (params.page.to_i - 1) : nil
