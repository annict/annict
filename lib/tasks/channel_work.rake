namespace :channel_work do
  task update: :environment do
    User.find_each do |user|
      puts "user_id: #{user.id} のchannel_workを更新しました。"
      ChannelWorkService.new(user).update
    end
  end
end
