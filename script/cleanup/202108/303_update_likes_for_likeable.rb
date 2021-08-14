# frozen_string_literal: true

opt = OptionParser.new
params = {}
opt.on("-u VAL") { |v| params[:user_id] = v }
opt.parse!(ARGV)

user_id = params[:user_id]

likes = Like.where(likeable_id: nil, likeable_type: nil).preload(:recipient)
likes = likes.where(user_id: user_id) if user_id

likes.find_each(order: :desc) do |l|
  p "likes.id: #{l.id}"

  likeable = case l.recipient_type
  when "AnimeRecord", "WorkRecord", "EpisodeRecord"
    l.recipient.record
  else
    l.recipient
  end

  if likeable
    l.update_columns(likeable_id: likeable.id, likeable_type: likeable.class.name)
  end
end
