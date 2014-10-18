namespace :keen do
  task publish_batch: :environment do
    events = {
      first_checkins: [],
      checkins: [],
      follows: [],
      users: [],
      first_statuses: [],
      statuses: []
    }

    User.order(:id).find_each do |u|
      puts "user: #{u.id}"

      if u.first_checkin.present?
        events[:first_checkins] << FirstCheckinsEvent.properties(:create, u.first_checkin)
      end

      if u.first_status.present?
        events[:first_statuses] << FirstStatusesEvent.properties(:create, u.first_status)
      end

      events[:users] << UsersEvent.properties(:create, u)
    end

    Checkin.find_each do |c|
      puts "checkin: #{c.id}"
      events[:checkins] << CheckinsEvent.properties(:create, c)
    end

    Follow.find_each do |f|
      puts "follow: #{f.id}"
      events[:follows] << FollowsEvent.properties(:create, f)
    end

    Status.find_each do |s|
      puts "status: #{s.id}"
      events[:statuses] << StatusesEvent.properties(:create, s)
    end

    puts 'publish batch...'
    Keen.publish_batch(events)
  end
end
