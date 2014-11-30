module UserReceivable
  extend ActiveSupport::Concern

  included do
    def receiving?(channel)
      receptions.where(channel_id: channel.id).present?
    end

    def receive(channel)
      unless receiving?(channel)
        receptions.create(channel: channel)
        ChannelWorksCreatingWorker.perform_async(id, channel.id)
      end
    end

    def unreceive(channel)
      reception = receptions.where(channel_id: channel.id).first

      if reception.present?
        reception.destroy
        ChannelWorksDestroyingWorker.perform_async(id, channel.id)
      end
    end
  end
end
