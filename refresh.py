from selenium import webdriver
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.common.keys import Keys
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC

# Path to the WebDriver executable (download from Selenium website)
#driver_path = '/path/to/webdriver/executable'

chrome_options = Options()
chrome_options.add_argument('--headless')  # Run in headless mode
chrome_options.add_argument('--disable-gpu')  # Disable GPU acceleration

# Initialize the WebDriver for Chrome
#driver = webdriver.Chrome(executable_path=driver_path)
driver = webdriver.Chrome(options=chrome_options)

# Open the Shiny app URL
driver.get('http://localhost:8090/?tab=dashboard_tab')

# Wait for specific element to be present
try:
    element = WebDriverWait(driver, 3600).until(
        EC.presence_of_element_located((By.ID, "httpvendor"))
    )
    print("Element found:", element.text)
except:
    print("Element not found within 10 seconds")

# Close the browser window
driver.quit()
