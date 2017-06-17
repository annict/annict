# frozen_string_literal: true

class SearchService
  attr_reader :q

  def initialize(q, scope: :published)
    @q = q
    @scope = scope
  end

  def series_list
    collection(Series).search(name_or_name_ro_or_name_en_cont: keyword).result
  end

  def works
    collection(Work).search(title_or_title_ro_or_title_en_or_title_kana_cont: keyword).result
  end

  def people
    collection(Person).search(name_or_name_en_or_name_kana_cont: keyword).result
  end

  def organizations
    collection(Organization).search(name_or_name_en_or_name_kana_cont: keyword).result
  end

  def characters
    collection(Character).search(name_or_name_en_or_name_kana_cont: keyword).result
  end

  private

  def keyword
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
