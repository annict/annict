class EpisodeDecorator < Draper::Decorator
  delegate_all

  def title_with_number
    if object.number.present? && object.title.present?
      "#{object.number}「#{object.title}」"
    elsif object.number.present? && object.title.blank?
      object.number
    elsif object.number.blank? && object.title.present?
      object.title
    end
  end
end
