# easyPreparation

## A Guide and Helper for Quick Preparations Before Worship

## 1. Automatic PPT Generation Program for Praise Titles and Lyrics
* Required Elements:
    - Church-specific background image

## 2. Easy Bulletin Creation Program (Incomplete)
* Required Elements:
    - Church-specific cover image
    - Bulletin content, etc.

![img.png](img.png)

## 3. Before Start 

### 3.1 Library Install

  ```shell
  apt update && apt install libreoffice && apt install Ghostscript
   # if this is not work, link the symbolic
  ln -s /Applications/LibreOffice.app/Contents/MacOS/soffice /usr/local/bin/libreoffice

  ```

### 3.2 Set Environment Variable(for ui build)

```shell
APP_ENV=dev
```


## 4. Info Size

* You can control the PDF size to change its ratio
* however, this is determined entirely by the Figma size.

```
- mac always follow 16:10 ratio

# 16:9
  width : 323.33,
  height : 181.0
  
# apply 16:9 size -> current size
  width : 1409.0,
  height : 792.5
  
# 16:10
  width : 323.33,
  height : 202.0
  
# apply 16:10 size -> current size
  width : 1409.0,
  height : 880.6
  
# A4 size
  width : 297.0,
  height : 210.0
 
# apply A4 size -> current size
  width : 1409.0,
  height : 996.0
 
  
  
  
```

## 5. Reference Repository

```
* figma(editable design tool):
      bulletin
          for print
              - background image(png)
          for presentation 
              - background template(png)
              
            
* Google Drive: 
      bulletin
          for presentation 
              - hymn lylics (pdf)
              - responsive_reading (pdf) 
              
* GitGub Gist(minize pool):
      web font info
              - fontinfo (json) 
```

## etc reference
```
# about data.info

"-" -> No modification
"b_edit" -> Bible edit / Modifications related to Bible verses
"c_edit" -> Obj edit / Modifications to the center information based on the main view
"r_edit" -> Lead edit / Modifications to the right information based on the main view

```
