module UserReceivable
  extend ActiveSupport::Concern

  included do
    def receiving?(channel)
      receptions.where(channel_id: channel.id).present?
    end

    def receive(channel)
      unless receiving?(channel)
        receptions.create(channel: channel)
        delay.create_channel_works(self, channel)
      end
    end

    def unreceive(channel)
      reception = receptions.where(channel_id: channel.id).first

      if reception.present?
        reception.destroy
        delay.destroy_channel_works(self, channel)
      end
    end

    private

    def create_channel_works(user, channel)
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

    def destroy_channel_works(user, channel)
      if user.present? && channel.present?
        user.works.wanna_watch_and_watching.each do |work|
          channel_work = user.channel_works.find_by(work_id: work.id, channel_id: channel.id)
          channel_work.destroy if channel_work.present?
        end
      end
    end
  end
end
