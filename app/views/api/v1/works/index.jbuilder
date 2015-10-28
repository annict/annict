@works.each do |work|
  json.partial! "/api/v1/works/work", work: work
end
