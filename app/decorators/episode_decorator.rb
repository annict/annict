class EpisodeDecorator < Draper::Decorator
  delegate_all

  def title_with_number
    if object.number.present?
      if object.title.present?
        "#{object.number}「#{object.title}」"
      else
        object.number
      end
    else
      object.title
    end
  end
end
