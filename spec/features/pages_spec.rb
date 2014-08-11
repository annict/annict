require 'spec_helper'

describe 'Annictについて' do
  let!(:work) { create(:work, :with_item) }

  context '未ログイン時' do
    before do
      visit '/about'
    end

    it 'アクセスするとページが表示される' do
      expect(page).to have_content('管理者について')
    end
  end
end

describe 'プライバシーポリシー' do
  before do
    visit '/privacy'
  end

  it 'アクセスするとページが表示される' do
    expect(page).to have_content('アニメ視聴記録によるコミュニケーション')
  end
end

describe '利用規約' do
  before do
    visit '/terms'
  end

  it 'アクセスするとページが表示される' do
    expect(page).to have_content('本利用規約')
  end
end
