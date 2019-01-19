# frozen_string_literal: true

module PvDecorator
  def db_detail_link(options = {})
    name = options.delete(:name).presence || id
    h.link_to(name, h.edit_db_pv_path(self), options)
  end

  def to_values
    model.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = send(field)
      hash
    end
  end
end
