# frozen_string_literal: true

opt = OptionParser.new
params = {}
opt.on("-u VAL") { |v| params[:user_id] = v }
opt.parse!(ARGV)

user_id = params[:user_id]

activities = Activity.where(itemable_id: nil, itemable_type: nil).preload(:itemable)
activities = activities.where(user_id: user_id) if user_id

activities.find_each(order: :desc) do |a|
  p "activities.id: #{a.id}"

  itemable = case a.trackable_type
  when "Status"
    a.itemable
  else
    a.itemable.record
  end

  if itemable
    a.update_columns(itemable_id: itemable.id, itemable_type: itemable.class.name)
  end
end