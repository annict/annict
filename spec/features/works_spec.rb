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

describe 'クールごとの作品一覧ページ' do
  let!(:season1) { create(:season) }
  let!(:season2) { create(:season) }
  let!(:work1) { create(:work, :with_item, season: season1) }
  let!(:work2) { create(:work, :with_item, season: season2) }

  before do
    visit "/works/#{season1.slug}"
  end

  it 'クールごとの作品が表示されること' do
    expect(page).to have_content(work1.title)
    expect(page).to_not have_content(work2.title)
  end
end

describe '人気の作品一覧ページ' do
  let!(:work) { create(:work, :with_item) }

  before do
    visit '/works/popular'
  end

  it '人気の作品が表示されること' do
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
