# frozen_string_literal: true

json.results @results do |result|
  json.resource_type result._type

  json.resource do
    case result._type
    when "work"
      json.call(result._source, :id, :title)
    when "person", "organization"
      json.call(result._source, :id, :name)
    end
  end
end
