# frozen_string_literal: true

[
  ChannelGroup,
  Channel,
  NumberFormat,
  Prefecture
].each do |model_class|
  file_path = "#{Dir.pwd}/db/data/csv/#{model_class.table_name}.csv"

  model_class.transaction do
    CSV.foreach(file_path, headers: true) do |row|
      record = model_class.where(id: row["id"]).first_or_create!(row.to_h.except("id"))
      puts "#{model_class}: #{record.id}"
    end
  end
end

ActiveRecord::Base.connection.execute("ALTER SEQUENCE channels_id_seq RESTART WITH #{Channel.last.id + 1};")
ActiveRecord::Base.connection.execute("ALTER SEQUENCE channel_groups_id_seq RESTART WITH #{ChannelGroup.last.id + 1};")
