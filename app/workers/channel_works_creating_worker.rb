class ChannelWorksCreatingWorker
  include Sidekiq::Worker

  def perform(user_id, channel_id)
    user = User.find(user_id)
    channel = Channel.find(channel_id)

    if user.present? && channel.present?
      user.works.wanna_watch_and_watching.each do |work|
        conditions =
            !user.channel_works.exists?(work_id: work.id) &&
            work.channels.present? &&
            work.channels.exists?(id: channel.id)

        if conditions
          user.channel_works.create(work: work, channel: channel)
        end
      end
    end
  end
end
