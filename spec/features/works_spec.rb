require 'spec_helper'

describe '放送中の作品一覧ページ' do
  let!(:work) { create(:work, :on_air, :with_item) }

  before do
    visit '/works'
  end

  it '放送中の作品が表示されること' do
    expect(page).to have_content(work.title)
  end
end

describe '作品検索ページ' do
  let!(:work) { create(:work, :on_air, :with_item) }

  before do
    visit '/works/search'

    within("#work_search") do
      fill_in 'q[title_cont]', with: work.title
    end
    click_button '検索'
  end

  it '検索結果に該当の作品が表示されること' do
    expect(find('.works-list')).to have_content(work.title)
  end
end