class ChannelWorksDestroyingWorker
  include Sidekiq::Worker

  def perform(user_id, channel_id)
    user = User.find(user_id)
    channel = Channel.find(channel_id)

    if user.present? && channel.present?
      user.works.wanna_watch_and_watching.each do |work|
        channel_work = user.channel_works.find_by(work_id: work.id, channel_id: channel.id)
        channel_work.destroy if channel_work.present?
      end
    end
  end
end
