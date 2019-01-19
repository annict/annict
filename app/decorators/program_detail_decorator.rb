# frozen_string_literal: true

module ProgramDetailDecorator
  def db_detail_link(options = {})
    name = options.delete(:name).presence || id
    h.link_to(name, h.edit_db_program_detail_path(self), options)
  end

  def to_values
    model.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = case field
      when :channel_id
        Channel.find(send(field)).name
      when :locale
        send(:locale_text)
      when :unique_id
        h.link_to unique_id, url, target: "_blank"
      else
        send(field)
      end

      hash
    end
  end
end
