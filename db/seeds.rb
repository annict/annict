# frozen_string_literal: true

[
  ChannelGroup,
  Channel,
  NumberFormat,
  Prefecture,
  SeasonModel,
  Tip,
  Work,
  Episode
].each do |model_class|
  file_path = "#{Dir.pwd}/db/data/csv/#{model_class.table_name}.csv"

  CSV.foreach(file_path, headers: true) do |row|
    record = model_class.where(id: row["id"]).first_or_create!(row.to_h.except("id"))
    puts "#{model_class}: #{record.id}"
  end
end
