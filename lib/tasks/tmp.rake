namespace :tmp do
  task change_share_checkin: :environment do
    User.find_each do |u|
      provider = u.providers.order(:id).first

      case provider.name
      when "twitter"
        u.setting.update_column(:share_record_to_twitter, u.share_checkin)
      when "facebook"
        u.setting.update_column(:share_record_to_facebook, u.share_checkin)
      end

      puts "user_id: #{u.id}"
    end
  end
end
