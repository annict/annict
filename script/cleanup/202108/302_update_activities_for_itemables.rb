# frozen_string_literal: true

opt = OptionParser.new
params = {}
opt.on("-u VAL") { |v| params[:user_id] = v }
opt.on("-f VAL") { |v| params[:from] = v }
opt.parse!(ARGV)

user_id = params[:user_id]
from = params[:from]

target_activities = Activity.where(itemable_id: nil, itemable_type: nil).preload(:itemable)
target_activities = target_activities.where(user_id: user_id) if user_id
target_activities = target_activities.after(from, field: :updated_at) if from

target_activities.find_in_batches(batch_size: 2_000, order: :desc) do |activities|
  attributes = activities.map do |activity|
    itemable_id, itemable_type = case activity.trackable_type
    when "Status"
      [activity.itemable_id, activity.itemable_type]
    else
      itemable = activity.itemable
      itemable&.record_id ? [itemable.record_id, "Record"] : [nil, nil]
    end

    activity.attributes.merge(
      "itemable_id" => itemable_id,
      "itemable_type" => itemable_type
    )
  end

  result = Activity.upsert_all(attributes)
  puts "upserted: #{result.rows.first(3)}..."
end
