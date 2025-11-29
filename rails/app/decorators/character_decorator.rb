# typed: false
# frozen_string_literal: true

module CharacterDecorator
  include RootResourceDecoratorCommon

  def name_link
    link_to local_name, character_path(self)
  end

  def name_with_series
    return local_name if series.blank?
    series_text = I18n.t("noun.series_with_name", series_name: series.local_name)
    "#{local_name} (#{series_text})"
  end

  def grid_description(cast)
    "CV: #{cast.person.decorate.name_link}"
  end

  def to_values
    self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = send(field)
      hash
    end
  end

  def db_detail_link(options = {})
    name = options.delete(:name).presence || self.name
    path = db_edit_character_path(self)
    link_to name, path, options
  end
end
