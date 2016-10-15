# frozen_string_literal: true

json.resources @characters do |character|
  json.id character.id
  json.text character.name
end
