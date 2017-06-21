# frozen_string_literal: true

class SearchService
  attr_reader :q

  def initialize(q, scope: :published)
    @q = q
    @scope = scope
  end

  def series_list
    collection(Series).search(name_or_name_ro_or_name_en_cont_any: keywords).result
  end

  def works
    collection(Work).search(title_or_title_ro_or_title_en_or_title_kana_cont_any: keywords).result
  end

  def people
    collection(Person).search(name_or_name_en_or_name_kana_cont_any: keywords).result
  end

  def organizations
    collection(Organization).search(name_or_name_en_or_name_kana_cont_any: keywords).result
  end

  def characters
    collection(Character).search(name_or_name_en_or_name_kana_cont_any: keywords).result
  end

  private

  def keywords
    [@q, keyword_hira]
  end

  def keyword_hira
    Moji.kata_to_hira(@q)
  end

  def collection(model)
    case @scope
    when :all
      model
    else
      model.published
    end
  end
end
