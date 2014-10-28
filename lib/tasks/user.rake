namespace :user do
  task update_finished_tips: :environment do
    status_tip = Tip.find_by(partial_name: 'status')
    channel_tip = Tip.find_by(partial_name: 'channel')
    checkin_tip = Tip.find_by(partial_name: 'checkin')

    User.find_each do |u|
      puts "user: #{u.username}"

      if u.statuses.present?
        u.finished_tips.create(tip: status_tip)
      end

      if u.channels.present?
        u.finished_tips.create(tip: channel_tip)
      end

      if u.checkins.present?
        u.finished_tips.create(tip: checkin_tip)
      end
    end
  end
end
