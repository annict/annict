# frozen_string_literal: true

json.works @works do |work|
  json.partial!("/api/v1/works/work", work: work, params: @params, field_prefix: "")
end
