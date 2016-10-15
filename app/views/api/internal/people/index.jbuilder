# frozen_string_literal: true

json.resources @people do |person|
  json.id person.id
  json.text person.name
end
