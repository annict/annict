# frozen_string_literal: true

json.works @works do |work|
  json.call(work, :id, :title)
end

json.people @people do |person|
  json.call(person, :id, :name)
end

json.organizations @organizations do |organization|
  json.call(organization, :id, :name)
end
