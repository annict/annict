namespace :recommendable do
  task update: :environment do
    r = Redis.new
    keys = r.keys('recommendable:*')
    r.del(keys) if keys.present?

    User.find_each do |user|
      watches  = [:wanna_watch, :watching, :watched]
      user.latest_statuses.each do |status|
        if watches.include?(status.kind.to_sym)
          user.like(status.work)
        else
          user.dislike(status.work)
        end
      end

      puts "user##{user.id} updated."
    end
  end
end