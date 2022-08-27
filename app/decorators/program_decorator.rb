# frozen_string_literal: true

module ProgramDecorator
  def name
    "#{channel.name} #{display_time(started_at)}~"
  end

  def db_detail_link(options = {})
    name = options.delete(:name).presence || "##{id}"
    link_to(name, db_edit_program_path(self), options)
  end

  def to_values
    self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = case field
      when :channel_id
        Channel.find(send(field)).name
      when :locale
        send(:locale_text)
      when :unique_id
        link_to unique_id, url, target: "_blank"
      else
        send(field)
      end

      hash
    end
  end
end
