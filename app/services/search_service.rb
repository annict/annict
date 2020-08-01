# frozen_string_literal: true

class SearchService
  attr_reader :q

  def initialize(q, scope: :published)
    @q = q
    @scope = scope
  end

  def series_list
    collection(Series).ransack(name_or_name_ro_or_name_en_cont_any: keywords).result
  end

  def works
    collection(Anime).ransack(title_or_title_en_or_title_kana_or_title_alter_or_title_alter_en_cont_any: keywords).result
  end

  def people
    collection(Person).ransack(name_or_name_en_or_name_kana_cont_any: keywords).result
  end

  def organizations
    collection(Organization).ransack(name_or_name_en_or_name_kana_cont_any: keywords).result
  end

  def characters
    collection(Character).ransack(name_or_name_en_or_name_kana_cont_any: keywords).result
  end

  private

  def keywords
    [@q, keyword_hira]
  end

  def keyword_hira
    return "" if @q.blank?

    Moji.kata_to_hira(@q)
  end

  def collection(model)
    case @scope
    when :all
      model
    else
      model.only_kept
    end
  end
end
